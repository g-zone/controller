# usage

The monitor can be configured at run time via command line options as follows:

./monitor [ {id}:{key}:{value} [ {id}:{key}:{value} ... ] ]

{id} is any ranfom string used to group {key}:{value} pairs belonging to same host
{key} can be any of: Host, PingPort, CmdOutPort, CmdInPort

Any number of hosts definitions can be added.

Default values are assumed for any {key} that is not entered on the command line. 

./node {server name} 

All other parameters assume default values (see "constants.go" in the parent directory)
