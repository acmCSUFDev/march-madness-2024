{
  description = "A very basic flake";

	inputs = {
		nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
		flake-utils.url = "github:numtide/flake-utils";

		gomod2nix.url = "github:nix-community/gomod2nix";
		gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
		gomod2nix.inputs.flake-utils.follows = "flake-utils";

		prettier-gohtml-nix.url = "github:diamondburned/prettier-gohtml-nix";
		prettier-gohtml-nix.inputs.nixpkgs.follows = "nixpkgs";
		prettier-gohtml-nix.inputs.flake-utils.follows = "flake-utils";
	};

	outputs =
		{ self, nixpkgs, flake-utils, gomod2nix, prettier-gohtml-nix }:

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
						# Go Tools.
						go
						gopls
						go-tools
						gomod2nix.packages.${system}.default
						sqlc
	
						# Web/JavaScript Tools.
						dart-sass
						prettier-gohtml-nix.packages.${system}.default
	
						# Python Tools.
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
