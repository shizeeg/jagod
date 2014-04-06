jagod
=====

Jabber Go Daemon is a XMPP/Jabber bot.

The project in heavy development at the moment. Use at your own risk.

The basic command syntax:

<prefix><section> <command> [parameters]

For example:

!user ping jago

<section> is any string at the moment so you can double the prefix like that:

!! ping jago

Parameter can be nick (in conference) or jabber-server (domain). Without parameter ping requester.

Some commands:

---------------------------------------------------------------------
XEP-0045: Multi-User Chat

!! join <conf@conference.server.tld>[/nick]
password

Joins the conference with optional nick and password.
---------------------------------------------------------------------
XEP-0199: XMPP Ping <http://xmpp.org/extensions/xep-0199.html>

!! ping [nick]|[server.tld]

ping server or occupant in conference.

---------------------------------------------------------------------
XEP-0202: Entity Time <http://xmpp.org/extensions/xep-0202.html>

!! version [nick]|[server.tld]

send iq:version request to nick or server.

---------------------------------------------------------------------
XEP-0249: Direct MUC Invitations <http://xmpp.org/extensions/xep-0249.html>

!! invite <nick>
[reason]

WARNING: This XEP is partially supported!

Now it just sends the actual invite. No checks for compatibility and no response implemented.
reason seems like do not work (why?).

---------------------------------------------------------------------
Yandex Dictionary (Yandex API key required!)

!! tr <lang1>-<lang2> <word>

Translate <word> from <lang1> to <lang2>
To see possible language pairs use:

!! tr help

---------------------------------------------------------------------
Yandex Spell

!! spell <text>

spell-check <text> using yandex spellchecker. Now supports english, russian &
ukrainian languages. You can set desired languages in config.

---------------------------------------------------------------------
Automatically unescape urls in messsages. You can add to config:

[cmd]
disable-turn-url=true 

To disable this functionality.

---------------------------------------------------------------------
Automatically translates cyrillic text typed in wrong layout. For example:

<shizeeg> ghbdtn? vbh!
<jago> привет, мир!

Disabled by default. You can add to config:

[cmd]
disable-turn=false

To enable this functionality.
---------------------------------------------------------------------

Official support conference: <xmpp:jagod@conference.jabber.org?join>
Development room: <xmpp:dev@conference.net45.uk.to?join> // no connection from jabber.ru/xmpp.ru at the moment, sorry.

---------------------------------------------------------------------

Many thanks goes to:

Adam Langley - for wonderful github.com/agl/xmpp lib.
speter       - for nice <http://code.google.com/p/gcfg> config parser.

and my CONTRIBUTORS!
