#!/usr/bin/nu

protoc -I=proto_src --go_out=plugins/p_seer_node_fleet/ proto_src/scraper.proto

