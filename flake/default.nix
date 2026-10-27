{
  imports = [
    ./devShells.nix
    ./git-hooks-nix.nix
    ./packages.nix
    ./treefmt.nix
  ];

  perSystem = {pkgs, ...}: {
    _module.args.goToolchain = pkgs.go_1_27;
  };
}
