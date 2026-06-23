{inputs, ...}: {
  flake.homeManagerModules.switchyard = import ./hm-module.nix inputs.self;

  perSystem = {
    pkgs,
    self',
    ...
  }: {
    packages = {
      default = self'.packages.switchyard;

      browser-setup = pkgs.python3Packages.buildPythonApplication {
        pname = "browser-setup";
        version = "dev";
        src = ../browser-setup;
        pyproject = true;
        build-system = with pkgs.python3.pkgs; [setuptools];
      };

      switchyard = pkgs.callPackage ../package.nix {};
    };
  };
}
