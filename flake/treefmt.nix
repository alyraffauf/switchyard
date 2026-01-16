_: {
  perSystem = _: {
    treefmt.config = {
      settings.global.excludes = ["vendor/*"];

      programs = {
        alejandra.enable = true;
        deadnix.enable = true;
        gofmt.enable = true;
        prettier.enable = true;
        shellcheck.enable = true;
        shfmt.enable = true;
        statix.enable = true;
      };
    };
  };
}
