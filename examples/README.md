# usage

The monitor can be configured at run time via command line options as follows:

<tt>monitor [ &lt;id&gt;:&lt;key&gt;:&lt;value&gt; [ &lt;id&gt;:&lt;key&gt;:&lt;value&gt; ... ] ]</tt>

<tt>&lt;id&gt;</tt> is any random string used to group <tt>&lt;key&gt;:&lt;value&gt;</tt> pairs belonging to same host.
<tt>&lt;key&gt;</tt> can be any of: <tt>Host, PingPort, CmdOutPort, CmdInPort</tt>

Any number of hosts definitions can be added.

Default values are assumed for any &lt;key&gt; that is not entered on the command line. 

</tt>node &lt;server name&gt;</tt>

All other parameters assume default values (see "constants.go" in the parent directory)
