alias is a simple wrapper for the function builtin, 
which creates a function wrapping a command. 
It has similar syntax to POSIX shell alias. 
For other uses, it is recommended to define a function.

fish marks functions that have been created by alias by 
including the command used to create them in the function 
description. You can list alias-created functions by running alias without 
arguments. They must be erased using functions -e.

NAME is the name of the alias

um dois tres quatro

DEFINITION is the actual command to execute. 
The string $argv will be appended.

You cannot create an alias to a function with the same name.
 Note that spaces need to be escaped in the call to alias just like at the command line, 
 even inside quoted parts.
