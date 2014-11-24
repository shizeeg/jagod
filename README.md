jagod
=====

Jabber Go Daemon is a XMPP/Jabber bot.

The project in heavy development at the moment. Use at your own risk.

The basic command syntax:

<pre>
&lt;prefix&gt;&lt;section&gt; &lt;command&gt; [parameters]
</pre>

For example:
<pre>
!user ping jago
</pre>
&lt;section&gt; is any string at the moment so you can double the prefix like that:
<pre>
!! ping jago
</pre>

Parameter can be nick (in conference) or jabber-server (domain). Without parameter ping requester.

Some commands:
---------------------------------------------------------------------
XEP-0045: Multi-User Chat
<pre>
!! join &lt;conf@conference.server.tld&gt;[/nick]
[password]
</pre>

Joins the conference with optional nick and password.

<pre>
!! leave [confjid]
[status/reason]
</pre>

If not specified, leaves the current conference with optional reason message.
You can also use "quit", "exit" or "part" instead.

---------------------------------------------------------------------
XEP-0092: Software Version <http://xmpp.org/extensions/xep-0092.html>

<pre>
!! version [nick]|[server.tld]
</pre>

Send iq:version request to nick or server.

---------------------------------------------------------------------
XEP-0199: XMPP Ping <http://xmpp.org/extensions/xep-0199.html>

<pre>
!! ping [nick]|[server.tld]
</pre>

Ping server or occupant in conference.

---------------------------------------------------------------------
XEP-0202: Entity Time <http://xmpp.org/extensions/xep-0202.html>

<pre>
!! time [nick]|[server.tld]
</pre>

Send urn:xmpp:time request to nick or server.

---------------------------------------------------------------------
XEP-0249: Direct MUC Invitations <http://xmpp.org/extensions/xep-0249.html>

<pre>
!! invite &lt;nick&gt;
[reason]
</pre>

WARNING: This XEP is partially supported!

Now it just sends the actual invite. No checks for compatibility and no response implemented.
reason seems like do not work (why?).

---------------------------------------------------------------------
Yandex Dictionary (Yandex API key required!)

<pre>
!! tr &lt;lang1&gt;-&lt;lang2&gt; &lt;word&gt;
</pre>

Translate &lt;word&gt; from &lt;lang1&gt; to &lt;lang2&gt;
To see possible language pairs use:

<pre>
!! tr help
</pre>

---------------------------------------------------------------------
Yandex Spell
<pre>
!! spell &lt;text&gt;
</pre>
spell-check <text> using yandex spellchecker. Now supports english, russian &
ukrainian languages. You can set desired languages in config.

---------------------------------------------------------------------
Google Search
<pre>
!! google &lt;request&gt;
</pre>

Search text in "request" on google.com

---------------------------------------------------------------------
Automatically unescape urls in messsages. You can add to config:

```cfg
[cmd]
disable-turn-url=true 
```

To disable this functionality.

---------------------------------------------------------------------
Automatically translates cyrillic text typed in wrong layout. For example:
<pre>
&lt;shizeeg&gt; ghbdtn? vbh!
&lt;jago&gt; привет, мир!
</pre>
Disabled by default. You can add to config:

```cfg
[cmd]
disable-turn=false
```

To enable this functionality.

---------------------------------------------------------------------
Show title of a webpage posted in MUC. For example:
<pre>
&lt;sh!zeeg&lt; look here is a new version of jagod released! https://github.com/shizeeg/jagod
&lt;jago&lt; Title: shizeeg/jagod · GitHub

---------------------------------------------------------------------

Development room: <xmpp:dev@conference.net45.tk?join>

---------------------------------------------------------------------

Many thanks goes to:

Go language team - for lovely programming language <http://golang.org/> <br />

@Adam Langley - for wonderful <https://github.com/agl/xmpp> lib. <br />
@speter       - for nice <http://code.google.com/p/gcfg> config parser. <br />

and my CONTRIBUTORS!
