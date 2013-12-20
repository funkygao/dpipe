=============
Kibana3 usage
=============

:Author: Gao Peng <funky.gao@gmail.com>
:Description: NA
:Revision: $Id$

.. contents:: Table Of Contents
.. section-numbering::

Search 
======

https://lucene.apache.org/core/3_5_0/queryparsersyntax.html

keynote
-------

- case sensitive

- wildcard supported

  * and ?

Boolean Operators
-----------------

- OR

  the default conjunction operator

- AND

- NOT

  cannot be used with just one term

use cases
---------

::

    uid:[55 TO 100000] AND _type:Cher
    name:"we are here"
    area:*e OR area:?l
    area:[aa TO zz]
    title:{Aida TO Carmen}  // Aida and Carmen not inclusive
    area:*e NOT de
    (area:ae OR area:nl) AND uid:[1 TO 1000000]

