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

					python3
					python3Packages.black
					pyright
				];
			};
		}
	);
}
