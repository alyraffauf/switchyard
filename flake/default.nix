{
  imports = [
    ./devShells.nix
    ./git-hooks-nix.nix
    ./packages.nix
    ./treefmt.nix
  ];

  perSystem = {pkgs, ...}: {
    # Pin the security release until nixpkgs provides it.
    _module.args.goToolchain = pkgs.go_1_27.overrideAttrs (finalAttrs: _: {
      version = "1.27.3";
      src = pkgs.fetchurl {
        url = "https://go.dev/dl/go${finalAttrs.version}.src.tar.gz";
        hash = "sha256-bndJA1zzYPBrJXXRYup535/ajvZxiJ5cN1+OUHd/Jrs=";
      };
    });
  };
}
