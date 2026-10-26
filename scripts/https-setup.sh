#!/usr/bin/env sh

set -eu

CONFIG_FILE="${CONFIG_FILE:-_datafiles/config.yaml}"
CURL_BIN="${CURL_BIN:-curl}"

if [ ! -f "$CONFIG_FILE" ]; then
	echo "Config file not found: $CONFIG_FILE" >&2
	exit 1
fi

read_yaml_value() {
	read_yaml_value_from_file "$CONFIG_FILE" "$1"
}

read_yaml_value_from_file() {
	awk -F': ' -v key="$2" '
		$1 ~ "^[[:space:]]*" key "$" {
			gsub(/^["'\'']|["'\'']$/, "", $2)
			print $2
			exit
		}
	' "$1"
}

yaml_key_exists_in_file() {
	awk -F': ' -v key="$2" '
		$1 ~ "^[[:space:]]*" key "$" {
			found = 1
			exit
		}
		END {
			exit(found ? 0 : 1)
		}
	' "$1"
}

canonicalize_path() {
	path=$1
	path_dir=$(dirname "$path")
	path_base=$(basename "$path")
	if abs_dir=$(cd "$path_dir" 2>/dev/null && pwd -P); then
		printf '%s/%s\n' "$abs_dir" "$path_base"
		return 0
	fi
	printf '%s\n' "$path"
}

json_escape() {
	printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

yaml_single_quote() {
	printf "'%s'" "$(printf '%s' "$1" | sed "s/'/''/g")"
}

base_url_without_trailing_slash() {
	url=$1
	while [ "${url%/}" != "$url" ]; do
		url=${url%/}
	done
	printf '%s' "$url"
}

normalize_hostname() {
	host=$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')
	host=${host#http://}
	host=${host#https://}
	host=${host%/}
	printf '%s' "$host"
}

is_public_hostname() {
	host=$(normalize_hostname "$1")

	case "$host" in
	"" | localhost | localhost.localdomain)
		return 1
		;;
	*:* | *" "* | */*)
		return 1
		;;
	*.local | *.internal | *.localhost)
		return 1
		;;
	esac

	case "$host" in
	*.*)
		;;
	*)
		return 1
		;;
	esac

	if printf '%s' "$host" | grep -Eq '^[0-9.]+$'; then
		return 1
	fi

	return 0
}

current_data_files=$(read_yaml_value DataFiles)
bundled_config_path=$(canonicalize_path "$CONFIG_FILE")
default_override_file="${current_data_files:-_datafiles/world/empty}/config-overrides.yaml"
override_file="${CONFIG_PATH:-$default_override_file}"
override_path=$(canonicalize_path "$override_file")

if [ "$override_path" = "$bundled_config_path" ]; then
	printf 'Ignoring CONFIG_PATH=%s because https-setup must not target the bundled base config.\n' "$override_file" >&2
	override_file=$default_override_file
fi

read_effective_yaml_value() {
	if [ -f "$override_file" ]; then
		if yaml_key_exists_in_file "$override_file" "$1"; then
			override_value=$(read_yaml_value_from_file "$override_file" "$1")
			printf '%s' "$override_value"
			return 0
		fi
	fi

	read_yaml_value "$1"
}

current_data_files=$(read_effective_yaml_value DataFiles)
current_https_cert=$(read_effective_yaml_value HttpsCertFile)
current_https_key=$(read_effective_yaml_value HttpsKeyFile)
current_web_domain=$(read_effective_yaml_value WebDomain)
current_http_port=$(read_effective_yaml_value HttpPort)
current_https_port=$(read_effective_yaml_value HttpsPort)
current_https_redirect=$(read_effective_yaml_value HttpsRedirect)
current_https_email=$(read_effective_yaml_value HttpsEmail)

https_cert_file=$current_https_cert
https_key_file=$current_https_key
web_domain=$current_web_domain
http_port=$current_http_port
https_port=$current_https_port
https_redirect=$current_https_redirect
https_email=$current_https_email
config_updates=""
override_snippet=""

printf 'Interactive HTTPS setup\n'
printf 'Bundled base config: %s\n' "$CONFIG_FILE"
printf 'Override target: %s\n\n' "$override_file"

printf 'Choose HTTPS mode:\n'
printf '  1) Manual certificate files\n'
printf '  2) Automatic Let'\''s Encrypt\n'
printf '  3) Disable HTTPS and use HTTP only\n'
printf 'Selection [1]: '
IFS= read -r mode_selection
mode_selection=${mode_selection:-1}
default_admin_base_url="http://localhost"

if [ "$mode_selection" = "2" ]; then
	default_admin_base_url="http://127.0.0.1"
	if [ -n "${current_http_port:-}" ] && [ "$current_http_port" != "80" ]; then
		default_admin_base_url="http://127.0.0.1:$current_http_port"
	fi
fi

case "$mode_selection" in
1)
	printf 'Certificate file path [%s]: ' "${current_https_cert:-server.crt}"
	IFS= read -r cert_input
	if [ -n "$cert_input" ]; then
		https_cert_file=$cert_input
	elif [ -z "$https_cert_file" ]; then
		https_cert_file=server.crt
	fi

	printf 'Private key file path [%s]: ' "${current_https_key:-server.key}"
	IFS= read -r key_input
	if [ -n "$key_input" ]; then
		https_key_file=$key_input
	elif [ -z "$https_key_file" ]; then
		https_key_file=server.key
	fi

	printf 'HTTP port [%s]: ' "${current_http_port:-80}"
	IFS= read -r http_port_input
	if [ -n "$http_port_input" ]; then
		http_port=$http_port_input
	elif [ -z "$http_port" ]; then
		http_port=80
	fi

	printf 'HTTPS port [%s]: ' "${current_https_port:-443}"
	IFS= read -r https_port_input
	if [ -n "$https_port_input" ]; then
		https_port=$https_port_input
	elif [ -z "$https_port" ]; then
		https_port=443
	fi

	printf 'Redirect HTTP to HTTPS? [%s]: ' "${current_https_redirect:-true}"
	IFS= read -r redirect_input
	if [ -n "$redirect_input" ]; then
		https_redirect=$redirect_input
	else
		https_redirect=${current_https_redirect:-true}
	fi

	config_updates=$(
		cat <<EOF
"FilePaths.HttpsCertFile":"$(json_escape "$https_cert_file")"
,"FilePaths.HttpsKeyFile":"$(json_escape "$https_key_file")"
,"Network.HttpPort":"$(json_escape "$http_port")"
,"Network.HttpsPort":"$(json_escape "$https_port")"
,"Network.HttpsRedirect":"$(json_escape "$https_redirect")"
EOF
	)

	override_snippet=$(
		cat <<EOF
FilePaths:
  HttpsCertFile: $(yaml_single_quote "$https_cert_file")
  HttpsKeyFile: $(yaml_single_quote "$https_key_file")
Network:
  HttpPort: $http_port
  HttpsPort: $https_port
  HttpsRedirect: $https_redirect
EOF
	)
	;;
2)
	if is_public_hostname "$current_web_domain"; then
		printf 'Public hostname [%s]: ' "$current_web_domain"
	else
		printf 'Public hostname (for example play.example.com): '
	fi
	IFS= read -r web_domain_input
	if [ -n "$web_domain_input" ]; then
		web_domain=$(normalize_hostname "$web_domain_input")
	fi

	if ! is_public_hostname "$web_domain"; then
		echo "Automatic HTTPS requires a public hostname like play.example.com, not localhost, a private-only name, or a raw IP address." >&2
		exit 1
	fi

	printf 'Contact email for Let'\''s Encrypt notices [%s]: ' "${current_https_email:-optional}"
	IFS= read -r https_email_input
	if [ -n "$https_email_input" ]; then
		https_email=$https_email_input
	elif [ -z "${https_email:-}" ]; then
		https_email=""
	fi

	http_port=80
	https_port=443
	https_redirect=true
	https_cert_file=""
	https_key_file=""

	config_updates=$(
		cat <<EOF
"FilePaths.WebDomain":"$(json_escape "$web_domain")"
,"FilePaths.HttpsEmail":"$(json_escape "$https_email")"
,"FilePaths.HttpsCertFile":""
,"FilePaths.HttpsKeyFile":""
,"Network.HttpPort":"80"
,"Network.HttpsPort":"443"
,"Network.HttpsRedirect":"true"
EOF
	)

	override_snippet=$(
		cat <<EOF
FilePaths:
  WebDomain: "$(json_escape "$web_domain")"
  HttpsEmail: "$(json_escape "$https_email")"
  HttpsCertFile: ''
  HttpsKeyFile: ''
Network:
  HttpPort: 80
  HttpsPort: 443
  HttpsRedirect: true
EOF
	)
	;;
3)
	printf 'HTTP port [%s]: ' "${current_http_port:-80}"
	IFS= read -r http_port_input
	if [ -n "$http_port_input" ]; then
		http_port=$http_port_input
	elif [ -z "$http_port" ]; then
		http_port=80
	fi

	https_port=0
	https_redirect=false
	https_cert_file=""
	https_key_file=""
	https_email=""

	config_updates=$(
		cat <<EOF
"FilePaths.HttpsCertFile":""
,"FilePaths.HttpsKeyFile":""
,"FilePaths.HttpsEmail":""
,"Network.HttpPort":"$(json_escape "$http_port")"
,"Network.HttpsPort":"0"
,"Network.HttpsRedirect":"false"
EOF
	)

	override_snippet=$(
		cat <<EOF
FilePaths:
  HttpsCertFile: ''
  HttpsKeyFile: ''
  HttpsEmail: ''
Network:
  HttpPort: $http_port
  HttpsPort: 0
  HttpsRedirect: false
EOF
	)
	;;
*)
	echo "Unknown selection: $mode_selection" >&2
	exit 1
	;;
esac

printf '\nPlanned settings:\n'
if [ "$mode_selection" = "2" ]; then
	printf '  WebDomain: %s\n' "${web_domain:-<empty>}"
	printf '  HttpsEmail: %s\n' "${https_email:-<empty>}"
fi
printf '  HttpsCertFile: %s\n' "${https_cert_file:-<empty>}"
printf '  HttpsKeyFile: %s\n' "${https_key_file:-<empty>}"
printf '  HttpPort: %s\n' "$http_port"
printf '  HttpsPort: %s\n' "$https_port"
printf '  HttpsRedirect: %s\n' "$https_redirect"

if [ "$mode_selection" = "1" ]; then
	printf '\nBefore applying these settings:\n'
	printf '  - Make sure %s and %s exist and are readable by the server.\n' "$https_cert_file" "$https_key_file"
	printf '  - Open inbound TCP port %s if players should use HTTPS from the internet.\n' "$https_port"
elif [ "$mode_selection" = "2" ]; then
	printf '\nBefore applying these settings:\n'
	printf '  - Point DNS for %s at this server or forwarded public endpoint.\n' "$web_domain"
	printf '  - Make sure inbound TCP ports 80 and 443 reach GoMud.\n'
	printf '  - Leave HttpsCertFile and HttpsKeyFile empty so GoMud can request a certificate.\n'
elif [ "$mode_selection" = "3" ]; then
	printf '\nBefore applying these settings:\n'
	printf '  - This turns HTTPS off and clears certificate and Let'\''s Encrypt email settings.\n'
	printf '  - Plain HTTP will continue on port %s until you re-enable HTTPS.\n' "$http_port"
fi

printf '\nChoose how to apply these changes:\n'
printf '  1) PATCH a running GoMud server via /admin/api/v1/config (recommended)\n'
printf '  2) Print a config-overrides snippet for manual save\n'
printf 'Selection [1]: '
IFS= read -r apply_selection
apply_selection=${apply_selection:-1}

case "$apply_selection" in
1)
	printf 'Admin base URL [%s]: ' "$default_admin_base_url"
	IFS= read -r admin_base_url
	admin_base_url=${admin_base_url:-$default_admin_base_url}
	admin_base_url=$(base_url_without_trailing_slash "$admin_base_url")

	printf 'Admin username [admin]: '
	IFS= read -r admin_username
	admin_username=${admin_username:-admin}

	printf 'Admin password: '
	IFS= read -r admin_password

	printf '\nPATCH %s/admin/api/v1/config\n' "$admin_base_url"
	curl_status=0
	curl_http_code=$("$CURL_BIN" --silent --show-error --output /dev/null --write-out '%{http_code}' \
		-u "$admin_username:$admin_password" \
		-H 'Content-Type: application/json' \
		-X PATCH \
		--data "{$config_updates}" \
		"$admin_base_url/admin/api/v1/config") || curl_status=$?
	if [ "$curl_status" -ne 0 ]; then
		printf '\nFailed to apply settings through the admin API.\n' >&2
		printf 'GoMud is not reachable at %s.\n' "$admin_base_url" >&2
		printf 'If the server is already running, enter its current admin URL and try again.\n' >&2
		printf 'Otherwise, save the override snippet below and restart GoMud.\n\n' >&2
		printf 'Fallback: save the following override snippet to %s and restart GoMud:\n\n' "$override_file" >&2
		printf '%s\n' "$override_snippet" >&2
		exit 1
	fi
	case "$curl_http_code" in
	2??)
		;;
	401 | 403)
		printf '\nFailed to apply settings through the admin API.\n' >&2
		printf 'GoMud responded with HTTP %s.\n' "$curl_http_code" >&2
		printf 'Check that the admin username and password are correct and that the account has admin access.\n\n' >&2
		printf 'Fallback: save the following override snippet to %s and restart GoMud:\n\n' "$override_file" >&2
		printf '%s\n' "$override_snippet" >&2
		exit 1
		;;
	*)
		printf '\nFailed to apply settings through the admin API.\n' >&2
		printf 'GoMud responded with HTTP %s.\n' "$curl_http_code" >&2
		printf 'Check the server logs or save the override snippet below and restart GoMud manually.\n\n' >&2
		printf 'Fallback: save the following override snippet to %s and restart GoMud:\n\n' "$override_file" >&2
		printf '%s\n' "$override_snippet" >&2
		exit 1
		;;
	esac

	printf '\nHTTPS setup applied through the admin API.\n'
	printf 'Next steps:\n'
	case "$mode_selection" in
	1)
		printf '  1. Confirm %s and %s exist and are readable.\n' "$https_cert_file" "$https_key_file"
		printf '  2. Restart GoMud so it rebinds the updated HTTP/HTTPS listeners.\n'
		printf '  3. Open https://your-domain:%s/.\n' "$https_port"
		;;
	2)
		printf '  1. Confirm %s resolves to this server and ports 80/443 are reachable.\n' "$web_domain"
		printf '  2. Restart GoMud so it rebinds the updated HTTP/HTTPS listeners.\n'
		printf '  3. Review /admin/https/ if certificate issuance needs troubleshooting.\n'
		;;
	3)
		printf '  1. Restart GoMud so it rebinds the updated HTTP/HTTPS listeners.\n'
		printf '  2. Connect over plain HTTP on port %s.\n' "$http_port"
		;;
	esac
	;;
2)
	printf '\nSave the following override snippet to %s:\n\n' "$override_file"
	printf '%s\n' "$override_snippet"
	printf '\nNext steps:\n'
	case "$mode_selection" in
	1)
		printf '  1. Save the snippet to %s or set the same keys through the admin API.\n' "$override_file"
		printf '  2. Confirm %s and %s exist and are readable.\n' "$https_cert_file" "$https_key_file"
		printf '  3. Restart GoMud and open https://your-domain:%s/.\n' "$https_port"
		;;
	2)
		printf '  1. Save the snippet to %s or set the same keys through the admin API.\n' "$override_file"
		printf '  2. Confirm %s resolves to this server and ports 80/443 are reachable.\n' "$web_domain"
		printf '  3. Restart GoMud and review /admin/https/ if certificate issuance needs troubleshooting.\n'
		;;
	3)
		printf '  1. Save the snippet to %s or set the same keys through the admin API.\n' "$override_file"
		printf '  2. Restart GoMud and connect over plain HTTP on port %s.\n' "$http_port"
		;;
	esac
	;;
*)
	echo "Unknown selection: $apply_selection" >&2
	exit 1
	;;
esac
