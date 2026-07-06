{
  buildGoModule,
  lib,
  stdenv,
  vendorHash,
}:
buildGoModule {
  pname = "switchyard-sw";
  version = "dev";
  src = ../.;
  inherit vendorHash;
  subPackages = ["cmd/sw"];

  # Darwin's browser scanner calls NSWorkspace via cgo; Linux's is pure Go,
  # so CGO stays off there for a fully static binary.
  env.CGO_ENABLED = if stdenv.isDarwin then "1" else "0";

  ldflags = [
    "-s"
    "-w"
  ];

  meta = with lib; {
    description = "Switchyard's terminal browser picker";
    homepage = "https://github.com/alyraffauf/switchyard";
    license = licenses.gpl3Plus;
    platforms = platforms.unix;
    mainProgram = "sw";
  };
}
