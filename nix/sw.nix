{
  buildGoModule,
  lib,
  vendorHash,
}:
buildGoModule {
  pname = "switchyard-sw";
  version = "dev";
  src = ../.;
  inherit vendorHash;
  subPackages = ["cmd/sw"];
  env.CGO_ENABLED = "0";

  ldflags = [
    "-s"
    "-w"
  ];

  meta = with lib; {
    description = "Switchyard's terminal browser picker";
    homepage = "https://github.com/alyraffauf/switchyard";
    license = licenses.gpl3Plus;
    platforms = platforms.linux;
    mainProgram = "sw";
  };
}
