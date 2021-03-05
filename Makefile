proto-compile:
	protoc \
	    --js_out=import_style=commonjs:web/public/js/ \
	    --go_out=game/ \
	    --go_opt=paths=source_relative \
	    proto/websocket.proto
