{
  description = "Description for the project";

  inputs = {
    devenv-root = {
      url = "file+file:///dev/null";
      flake = false;
    };
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:cachix/devenv-nixpkgs/rolling";
    devenv.url = "github:cachix/devenv";
  };

  outputs = inputs@{ flake-parts, devenv-root, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.devenv.flakeModule
      ];
      systems = [ "x86_64-linux" "aarch64-linux" ];
      perSystem = { config, self', inputs', pkgs, lib, system, ... }: {
        #packages.default = pkgs.buildGoModule {
        #  pname = "mdbin";
        #  version = "v0.0.1";
        #  src = ./.;
        #  #vendorHash = lib.fakeHash;
        #  vendorHash = "sha256-enny+P5iY5X6GkOLtZpm3SvrG/laP5UnXBjAi9a7++8=";
        #};

        devenv.shells.default = {
          name = "mdbin";
          languages.go.enable = true;

          scripts = {
            get-go-version.exec = ''
              go version | grep -oP 'go\K[0-9.]+'
            '';
            run-go-maintanence.exec = ''
              go mod edit -go $(get-go-version)
              go mod tidy
            '';
            publish-image.exec = ''
              set -e
              git branch --show-current | grep -q main || { echo "Current branch is not main. aborting"; exit 1; }
              run-go-maintanence # if this changes anything, the next step will fail (intended behavior)
              git status --porcelain | grep -q . && { echo "Main branch is not clean. aborting"; exit 1; } || :
              git fetch origin main && git diff --quiet origin/main HEAD || echo "Local main branch differs from remote. aborting" && exit 1
              sudo docker build --build-arg GOVERSION=$(get-go-version)-alpine --platform linux/amd64,linux/arm64 -t lcwllmr/mdbin .
              sudo docker push lcwllmr/mdbin:$(git rev-parse --short HEAD)
            '';
          };
        };
      };
    };
}
