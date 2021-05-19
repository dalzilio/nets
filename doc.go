// Copyright (c) 2021 Silvano DAL ZILIO
//
// GNU Affero GPL v3

/*
Package nets defines a concrete type for Time Petri Nets and provides a Parser
for building Nets from the textual description format used in the Tina toolbox
(see below).

The net format

We support a very general subset of the description format for Time Petri nets
found in the Tina man pages (see
http://projects.laas.fr/tina/manuals/formats.html). We explain some of our
restrictions below. All the files successfully parsed or printed using the
library are valid .net file

A net is described by a series of declarations of places, transitions,
priorities  and/or notes, and an optional naming declaration for the net. The
net described is the superposition of these declarations. The grammar of .net
declarations is the following, in which nonterminals are bracketed by < .. >,
terminals are in upper case or quoted. Spaces, carriage return and tabs act as
separators.

Labels may be (optionally) assigned to places and transitions, but we do not
support the use of a "lb" declaration, for labels, that was only kept for
backward compatibility. We also do not support stopwatches and reset arcs.

Grammar

    .net                    ::= (<trdesc>|<pldesc>|<lbdesc>|<prdesc>|<ntdesc>|<netdesc>)*
    netdesc                 ::= ’net’ <net>
    trdesc                  ::= ’tr’ <transition> {":" <label>} {<interval>} {<tinput> -> <toutput>}
    pldesc                  ::= ’pl’ <place> {":" <label>} {(<marking>)}
    ntdesc                  ::= ’nt’ <note> (’0’|’1’) <annotation>
    prdesc                  ::= ’pr’ (<transition>)+ ("<"|">") (<transition>)+
    interval                ::= (’[’|’]’)INT’,’INT(’[’|’]’) | (’[’|’]’)INT’,’w[’
    tinput                  ::= <place>{<arc>}
    toutput                 ::= <place>{<normal_arc>}
    arc                     ::= <normal_arc> | <test_arc> | <inhibitor_arc> |
                                <stopwatch_arc> | <stopwatch-inhibitor_arc>
    normal_arc              ::= ’*’<weight>
    test_arc                ::= ’?’<weight>
    inhibitor_arc           ::= ’?-’<weight>
    weight, marking         ::= INT{’K’|’M’}
    net, place, transition,
    label, note, annotation ::= ANAME | ’{’QNAME’}’
    INT                     ::= unsigned integer
    ANAME                   ::= alphanumeric name, see Notes below
    QNAME                   ::= arbitrary name, see Notes below

Notes

Two forms are admitted for net, place and transition names:

     - ANAME : any non empty string of letters, digits, primes (’) and underscores (_)

     - ’{’QNAME’}’ : any chain between braces, and in which the three characters "{,}, or \" are escaped with a \

Empty lines and lines beginning with ’#’ are considered comments.

In any closed temporal interval [eft,lft], one must have eft <= lft.

Weight is optional for normal arcs, but mandatory for test and inhibitor arcs.

By default: transitions have temporal interval [0,w[; normal arcs have weight 1;
places have marking 0; and transitions have the empty label "{}"

When several labels are assigned to some node, only the last assigned is kept.

Simple example of .net file

This is a simple example of .net file. Note that it is possible to have several
declarations for the same object (place or transition); the end result is the
fusion of all these declarations.

     tr t1 p1 p2*2 -> p3 p4 p5
     tr t2 [0,2] p4 -> p2
     tr t3 : a p5 -> p2
     tr t3 p3 -> p3
     tr t4 [0,3] p3 -> p1
     pl p1 (1)
     pl p2 (2)

*/
package nets
