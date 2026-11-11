# Help for ~get~

The ~get~ command picks up an item on the ground and puts it in your backpack.

## Usage:

  ~get stick~  
  This would get a stick from the ground and put it in your backpack.  
  ~get stick from stash~  
  This would get a stick stashed in the area and put it in your backpack.  
  ~get all bob~  
  This would take everything from Bob's corpse (~get all from bob~ works too).  
  ~get stick bob~  
  This would take a stick from Bob's corpse.

A corpse holds what its owner dropped until it crumbles to dust. When it crumbles, everything still in it is destroyed.

A name counts the living first, then corpses from the newest. With a live rat and two rat corpses here, ~get all 2.rat~ takes everything from the newest corpse and ~get all 3.rat~ from the older (~rat#2~ works too); ~get all rat~ is refused, since the rat is alive. Your own name skips you, so ~get all~ and your name finds your newest corpse, unless someone here has a name that starts with yours (Bobby, for Bob): they come first, and your corpse is then ~get all 2.bob~. With no corpse of that name here, ~get all~ and a name takes everything on the floor, as plain ~get all~ does. Someone you cannot see is not counted.

Find out more about referring to items by name by typing ~help item-names~.