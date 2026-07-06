{inputs, ...}: {
  flake.homeManagerModules.switchyard = import ./hm-module.nix inputs.self;

  perSystem = {
    pkgs,
    self',
    ...
  }: let
    # Shared across both Go packages
    vendorHash = "sha256-Mkb8LSMUeKqOHxvHq7W1/7rxRr73Cx3L1Pf7+xjWgaE=";
  in {
    packages = {
      default = self'.packages.switchyard;

      browser-setup = pkgs.python3Packages.buildPythonApplication {
        pname = "browser-setup";
        version = "dev";
        src = ../browser-setup;
        pyproject = true;
        build-system = with pkgs.python3.pkgs; [setuptools];
      };

      # GTK GUI (cmd/switchyard).
      switchyard = pkgs.callPackage ../nix/switchyard.nix {inherit vendorHash;};

      # Terminal browser picker (cmd/sw).
      sw = pkgs.callPackage ../nix/sw.nix {inherit vendorHash;};
    };
  };
}
