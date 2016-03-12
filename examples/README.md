# usage

The monitor can be configured at run time via command line options as follows:

<tt>monitor [ {id}:{key}:{value} [ {id}:{key}:{value} ... ] ]</tt>

<b>{id}</b> is any random string used to group <b>{key}:{value}</b> pairs belonging to same host.
<b>{key}</b> can be any of: <tt>Host, PingPort, CmdOutPort, CmdInPort</tt>

Any number of hosts definitions can be added.

Default values are assumed for any <b>{key}</b> that is not entered on the command line. 

</tt>node {server name}</tt>

All other parameters assume default values (see "constants.go" in the parent directory)
