============================
engine pipeline architecture
============================

:Author: Gao Peng <funky.gao@gmail.com>
:Description: NA
:Revision: $Id$

.. contents:: Table Of Contents
.. section-numbering::

PipelinePack DataFlow
=====================

- Input/InputRunner

- Output/OutputRunner

- Filter/FilterRunner

- Router

::

    EngineConfig.(PoolSize of PipelinePack)
            |
            | is
            |
    InputRunner.InChan
            |
            |     +--------------------------------------------------+
    consume |     | Router.(PluginChanSize of PipelinePack)          |
            |     +--------------------------------------------------+
          Input         ^           |               |           ^
            |           |           | put           | put       |
            V           |           V               V           |
            +-----------+       OutputRunner   FilterRunner     |
              inject                |               |           |
                                    | consume       | consume   | inject
                                    V               V           |
                                 Output           +----------------+
                                                  |   Filter       |
                                                  +----------------+


