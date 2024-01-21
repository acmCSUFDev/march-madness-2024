{
  description = "A very basic flake";

	inputs = {
		nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
		flake-utils.url = "github:numtide/flake-utils";
		gomod2nix.url = "github:nix-community/gomod2nix";
		gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
		gomod2nix.inputs.flake-utils.follows = "flake-utils";
	};

	outputs =
		{ self, nixpkgs, flake-utils, gomod2nix }:

		flake-utils.lib.eachDefaultSystem (system:
			with nixpkgs.legacyPackages.${system};
			with gomod2nix.legacyPackages.${system};
			{
				packages.default = buildGoApplication {
					pname = "february-frenzy";
					version = "git";

					pwd = ./.;
					src = ./.;
					modules = ./gomod2nix.toml;

					preBuild = "go generate ./...";
					doCheck = false;

					subPackages = [
						"."
						"./cmd/competitionctl"
					];

					nativeBuildInputs = [
						sqlc
						dart-sass
					];
				};
				devShell = mkShell {
					packages = [
						go
						gopls
						go-tools
						(gomod2nix.legacyPackages.${system}).gomod2nix
	
						sqlc
	
						nodejs
						dart-sass
	
						python3
						python3Packages.black
						pyright
					];
	
					shellHook = ''
						python3 -m venv .venv
						source .venv/bin/activate
						export PATH="$PATH:$(git rev-parse --show-toplevel)/node_modules/.bin"
					'';
				};
			}
		);
}
