{
  description = "Flake for March Madness 2024";

	inputs = {
		nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
		flake-utils.url = "github:numtide/flake-utils";

		gomod2nix.url = "github:nix-community/gomod2nix";
		gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
		gomod2nix.inputs.flake-utils.follows = "flake-utils";

		poetry2nix.url = "github:nix-community/poetry2nix";
		poetry2nix.inputs.nixpkgs.follows = "nixpkgs";
		poetry2nix.inputs.flake-utils.follows = "flake-utils";

		prettier-gohtml-nix.url = "github:diamondburned/prettier-gohtml-nix";
		prettier-gohtml-nix.inputs.nixpkgs.follows = "nixpkgs";
		prettier-gohtml-nix.inputs.flake-utils.follows = "flake-utils";
	};

	outputs =
		{ self, nixpkgs, flake-utils, gomod2nix, poetry2nix, prettier-gohtml-nix }:

		flake-utils.lib.eachDefaultSystem (system:
			let
				pkgs = nixpkgs.legacyPackages.${system};
				python = pkgs.python3;
				packages = self.packages.${system};
			in

			with pkgs;
			with gomod2nix.legacyPackages.${system};
			with poetry2nix.lib.mkPoetry2Nix { inherit pkgs; };

			let
				problemsPoetry = mkPoetryApplication {
					pname = "march-madness-problems";
					doCheck = false;
					projectDir = self;
					preferWheels = true;
				};
			in

			{
				packages = {
					default = pkgs.runCommandLocal "march-madness" {
						nativeBuildInputs = [ makeWrapper ];
						buildInputs = with self.packages.${system}; [ server problems ];
						meta = with pkgs.lib; {
							description = "Executables for March Madness 2024 server and tools";
							homepage = "https://dev.acmcsuf.com/march-madness-2024";
							mainProgram = "march-madness";
						};
					} ''
						mkdir -p $out/bin

						# bin/competitionctl
						ln -s ${packages.server}/bin/competitionctl $out/bin/

						# bin/march-madness
						makeWrapper \
							${packages.server}/bin/march-madness-2024 \
							$out/bin/march-madness \
								--suffix PATH : ${packages.problems}/bin
					'';
					server = buildGoApplication {
						name = "march-madness-server";
						pwd = ./.;
						src = ./.;
						modules = ./gomod2nix.toml;
	
						doCheck = false;
						preBuild = "go generate ./...";
	
						subPackages = [
							"."
							"./cmd/competitionctl"
						];
	
						nativeBuildInputs = [
							sqlc
							dart-sass
						];
					};
					problems = problemsPoetry.dependencyEnv;
				};
				devShell = mkShell {
					inputsFrom = [ problemsPoetry ];

					packages = [
						# Go Tools.
						go
						gopls
						gotools
						go-tools
						gomod2nix.packages.${system}.default
						sqlc
	
						# Web/JavaScript Tools.
						deno
						dart-sass
						prettier-gohtml-nix.packages.${system}.default
	
						# Python Tools.
						python3
						python3Packages.black
						pyright
						poetry
					];

					DENO_NO_UPDATE_CHECK = "1"; # Useless in Nix.
				};
			}
		);
}
