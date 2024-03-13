{
  description = "A very basic flake";

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
					server = buildGoApplication {
						pname = "march-madness-server";
						version = "git";
	
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
					problems = runCommandLocal "march-madness-problems" {
						buildInputs = [ problemsPoetry.dependencyEnv ];
					} ''
						mkdir -p $out/bin

						cd "${problemsPoetry.dependencyEnv}/${problemsPoetry.python.sitePackages}"

						for problem in problems/*/__main__.py; do
							modulePath="$(dirname $problem)"
							binaryName="$(basename $modulePath)"
							module="''${modulePath//\//.}"
							script="$out/bin/problem-$binaryName"

							echo "#!${runtimeShell}" >> $script
							echo "exec ${problemsPoetry.dependencyEnv}/bin/python3 -m $module \"\$@\"" >> $script

							chmod +x $script
						done
					'';
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
