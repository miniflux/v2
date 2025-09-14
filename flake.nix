{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };
      in
      {
        devShells.default = pkgs.mkShell {
          PGHOST = "localhost";
          PGPORT = 5432;
          PGPASSWORD = "postgres";
          PGUSER = "postgres";
          PGDATABASE = "miniflux2";

          COCKROACH_URL = "postgresql://postgres:postgres@localhost:26257/miniflux2";
          COCKROACH_INSECURE = true;

          packages = with pkgs; [
            git

            gnumake
            foreman

            nil
            nixfmt-rfc-style

            nodePackages.prettier
            nodePackages.yaml-language-server
            nodePackages.vscode-langservers-extracted
            markdownlint-cli
            nodePackages.markdown-link-check
            marksman
            taplo

            go
            gopls
            go-tools
            gofumpt
            golangci-lint

            postgresql
            cockroachdb
            sqlite
            usql
          ];
        };
      }
    );
}
