// Copyright (c) 2021 Silvano DAL ZILIO
//
// GNU Affero GPL v3

// Package nets defines a concrete type for Time Petri Nets and provides a
// Parser for building Nets from the textual description format used in the Tina
// toolbox (see below).

package nets

// The net format
//
// The textual description format for Time Petri nets can be found here:
// http://projects.laas.fr/tina/manuals/formats.html.
//
// A net is described by a series of declarations of places, transitions and/or
// notes, and an optional naming declaration for the net. The net described is
// the superposition of these declarations. The grammar of .net declarations is
// the following, in which nonterminals are bracketed by < .. >, terminals are
// in upper case or quoted. Spaces, carriage return and tabs act as separators.
//
// Optionally, labels may be assigned to places and transitions. This should be
// preferably done within "tr" and "pl" declarations rather than using separate
// "lb" declarations. The later form ("lb") is kept for backward compatibility
// and might disappear in future releases.
//
// Grammar
//
//       .net                    ::= (<trdesc>|<pldesc>|<lbdesc>|<prdesc>|<ntdesc>|<netdesc>)*
//       netdesc                 ::= ’net’ <net>
//       trdesc                  ::= ’tr’ <transition> {":" <label>} {<interval>} {<tinput> -> <toutput>}
//       pldesc                  ::= ’pl’ <place> {":" <label>} {(<marking>)} {<pinput> -> <poutput>}
//       ntdesc                  ::= ’nt’ <note> (’0’|’1’) <annotation>
//       lbdesc                  ::= ’lb’ [<place>|<transition>] <label>
//       prdesc                  ::= ’pr’ (<transition>)+ ("<"|">") (<transition>)+
//       interval                        ::= (’[’|’]’)INT’,’INT(’[’|’]’) | (’[’|’]’)INT’,’w[’
//       tinput                  ::= <place>{<arc>}
//       toutput                 ::= <place>{<normal_arc>}
//       pinput                  ::= <transition>{<normal_arc>}
//       poutput                 ::= <transition>{arc}
//       arc                     ::= <normal_arc> | <test_arc> | <inhibitor_arc> |
//                                   <stopwatch_arc> | <stopwatch-inhibitor_arc>
//       normal_arc              ::= ’*’<weight>
//       test_arc                ::= ’?’<weight>
//       inhibitor_arc           ::= ’?-’<weight>
//       stopwatch_arc           ::= ’!’<weight>
//       stopwatch-inhibitor_arc ::= ’!-’<weight>
//       weight, marking         ::= INT{’K’|’M’}
//       net, place, transition, label, note, annotation ::= ANAME | ’{’QNAME’}’
//       INT                     ::= unsigned integer
//       ANAME                   ::= alphanumeric name, see Notes below
//       QNAME                   ::= arbitrary name, see Notes below
//
// Notes
//
// Two forms are admitted for net, place and transition names:
//
//      - ANAME : any non empty string of letters, digits, primes (’) and underscores (_)
//
//      - ’{’QNAME’}’ : any chain between braces, and in which characters in "{, }, \" are escaped with a \
//
// Empty lines and lines beginning with ’#’ are considered comments.
//
// In any closed temporal interval [eft,lft], one must have eft <= lft.
//
// Weight is optional for normal arcs, but mandatory for test and inhibitor arcs
//
// By default: transitions have temporal interval [0,w[; normal arcs have weight
// 1; places have marking 0; and transitions have the empty label "{}"
//
// When several labels are assigned to some node, only the last assigned is
// kept.
