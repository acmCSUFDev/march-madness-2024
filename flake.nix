{
  description = "A very basic flake";

	inputs = {
		nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
		flake-utils.url = "github:numtide/flake-utils";
	};

	outputs = { self, nixpkgs, flake-utils }: flake-utils.lib.eachDefaultSystem (system:
		with import nixpkgs { inherit system; };
		{
			devShell = mkShell {
				packages = [
					go
					gopls
					go-tools

					sqlc

					nodejs
					dart-sass
					(writeShellScriptBin "prettier" ''
						cd $(git rev-parse --show-toplevel)/server/frontend
						exec node_modules/.bin/prettier "$@"
					'')

					python3
					python3Packages.black
					pyright
				];

				shellHook = ''
					python3 -m venv .venv
					source .venv/bin/activate
				'';
			};
		}
	);
}
