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

buffer size of PipelinePack

- EngineConfig

  PoolSize 

- Input/InputRunner

  PluginChanSize 

- Output/OutputRunner

  PluginChanSize

- Filter/FilterRunner

  PluginChanSize 

- Router

  PluginChanSize

::


                        -------<--------+
                        |               |
                        V               | generate pool
       EngineConfig.inputRecycleChan    | recycling
            |           |               |
            | is        +------->-------+
            |
    InputRunner.inChan
            |
            |     +--------------------------------------------------------+
    consume |     |                     Router.inChan                      |
            |     +--------------------------------------------------------+
          Input         ^           |               |                   ^
            |           |           | put           | put               |
            V           |           V               V                   |
            +-----------+  OutputRunner.inChan   FilterRunner.inChan    |
              inject                |               |                   |
                                    | consume       | consume           | inject
                                    V               V                   |
                                 Output           +------------------------+
                                                  |         Filter         |
                                                  +------------------------+


