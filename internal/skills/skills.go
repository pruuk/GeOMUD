package skills

// Skills are identified by lowercase string ids (e.g. "cast", "dual-wield").
// The YAML datafiles in _datafiles/world/default/skills are the source of truth
// for which skills exist at runtime. Code that gates behavior on a specific
// skill references the id as a plain string literal.
