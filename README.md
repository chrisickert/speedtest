# speedtest

An (internet) speed test client written in Go that
* measures internet speed using Ookla's speedtest cli
* writes results into a PostgreSql database

Can be used to periodically (every 5 mins) do a speed test (e.g. on a Raspberry Pi) in order to validate your internet provider's performance.
