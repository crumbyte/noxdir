{
  description = "High-performance, cross-platform command-line tool for visualizing and exploring your file system usage";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            golangci-lint
          ];
        };

        packages.noxdir = pkgs.buildGoModule {
          pname = "noxdir";
          version = "0.10.0";
          src = ./.;
          vendorHash = "sha256-uRJP21bJ8NlJ0qOG81Gax9LJ+HdPfxLKj1Jjzbweync=";
          ldflags = [
            "-s"
            "-w"
          ];
          subPackages = [ "." ];
        };

        packages.default = self.packages.${system}.noxdir;
      }
    );
}
