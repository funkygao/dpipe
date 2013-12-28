Plugin
======

Each plugin actually consists of 2 parts:
* plugin itself
contains the plugin-specific behaviour and is provided by plugin developer.
* plugin runner
contains the shared(by type) behaviour and is provided by system.
    - InputRunner
    - DecoderRunner
    - FilterRunner
    - OutputRunner

When system starts a plugin, it
*   creates and configures a plugin instance of the appropirate type
*   creates a plugin runner
*   calls the Start() of the plugin runner
*   plugin runners then call the plugin's Run(), passing themselves and an additional PluginHelper object as arguments so the plugin code can use their exposed APIs to interact with system

### Types

*   Input
*   Decoder

    Convert raw data received by input into Message structs that can be
    processed by filters and outputs.

*   Filter

    Aggregation/roll-ups
    Simple event processing / windowed stream operations

*   Output
