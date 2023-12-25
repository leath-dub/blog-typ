{
  description = "Third Year Project blog";

  inputs = {
    templ.url = "github:a-h/templ";
  };

  outputs = inputs@{ self, nixpkgs, systems, ... }: let
    forAllSystems = fn:
      nixpkgs.lib.genAttrs (import systems) (system: fn {
        inherit system;
        pkgs = import nixpkgs { inherit system; };
      });
    templ = system: inputs.templ.packages.${system}.templ;
  in {
    packages = forAllSystems ({ pkgs, system }: with pkgs; {
      default = buildGoModule {
        name = "blog-typ";
        src = ./.;
        vendorHash = "sha256-A7hZP0luwu6P/LyuRzINjOTpmqYWsY95w7zRY/KsZR0=";
        preBuild = ''
          ${templ system}/bin/templ generate
        '';
      };
    });

    devShells = forAllSystems ({ pkgs, system }: with pkgs; {
      default = mkShell {
        buildInputs = [
          go
          (templ system)
        ];

        shellHook = ''
        exec ash
        '';
      };
    });
  };
}
