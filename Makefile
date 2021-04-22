proto-compile:
	protoc \
		--js_out=import_style=commonjs,binary:web/src/ \
		--go_out=game/ \
		--go_opt=paths=source_relative \
		proto/input.proto
