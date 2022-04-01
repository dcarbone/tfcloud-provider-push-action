# Terraform Cloud Provider Push Action
A basic action to push a custom [Terraform](https://www.terraform.io/) 
[Provider Plugin](https://www.hashicorp.com/blog/writing-custom-terraform-providers) to a Terraform Cloud private
registry.

## Key Links
1. [Publishing Private Providers to the Terraform Registry](https://www.terraform.io/cloud-docs/registry/publish-providers)
2. [GoReleaser](https://goreleaser.com/)
   * [GoReleaser Github Action](https://github.com/goreleaser/goreleaser-action)

# Setup
Before you can use this action, you must currently perform some manual set up steps in your Terraform Cloud instance.

## 1. Create Personal API Token
In your Terraform Cloud instance, go to the [Tokens](https://app.terraform.io/app/settings/tokens) page for your
user. *Note*: the link will only work if you're already logged in to Terraform Cloud.

Once there, click on the purple `Create API token` button.  Save this token in a secure location, such as a password
manager, as you'll need it in subsequent steps.

This token will be referred to below as the "personal" token.

## 2. Create a "Robots" API Token
In order for this action to push to your org's private registry, you must create an API token with `Manage Providers`
permissions.  This can either be a Team or User token.

Personally, I recommend creating a shared `Robots` team with a restricted set of permissions, and either creating a
token for that team or create a robot user, adding them to the team, and creating a token on that user.  Do whatever
makes sense for your org.

This token will be referred to below as the "robot" token.

## 3. GPG Key
Terraform requires that all providers and plugins are signed by registered GPG keys.  You must create a new GPG key
and register it with Terraform Cloud yourself.

#### Creating a GPG Key
Github has provided a great tutorial on how to create a GPG key here:
https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key.

A few notes:

1. Use `RSA` for both.
2. May as well make it `4096` bits in size.
3. Don't use this key for anything else.
4. It is a good idea to enforce an expiration on this key, but that does, of course, mean you'll need to create
   a way to generate and push a new key to Terraform Cloud and update all references to that key ID.

#### Registering a GPG Key
Once created, use [curl](https://www.man7.org/linux/man-pages/man1/curl.1.html) or [Postman](https://www.postman.com/)
or whatever you want to execute the 
[Add a GPG Key](https://www.terraform.io/cloud-docs/api-docs/private-registry/gpg-keys#add-a-gpg-key) HTTP request.

Save the response from `key-id` in a secure location.  You must be able to provide the `key-id` value when running
this action.

## 4. Release Assets Structure
Currently this action expects to be triggered by the creation of a Github Release with a specific list of attached
assets.

### File Naming
Each file must follow this naming structure:
```
terraform-provivder-{{ provider name }}_{{ semver }}_{{ * }}
```

### Required Files
There are three distinct types of required files:

1. Compressed binary artifact `.zip` file(s)
2. `SHA256SUMS`
3. `SHA256SUMS.sig`

#### Compressed Binaries
Per platform you wish to support, there must exist a `.zip` compressed binary asset attached to the release with the
following naming schema:

```
terraform-provider-{{ provider name }}_{{ semver }}_{{ os }}_{{ arch }}.zip
```

As an example:

```
terraform-provider-myprovider_0.1.0_linux_amd64.zip
```

The above will be parsed into the following components:

1. Provider name is `myprovider`
2. Version `0.1.0`
3. OS `linux`
4. Arch `amd64`

You _must_ create a file _per_ `provider : os : arch` combination!  Do not bundle multiple binaries into a single 
compressed artifact.

##### Suggested OS and Architecture Combinations
* freebsd
  * amd64
  * arm64
* linux
  * amd64
  * arm64
* darwin
  * amd64
  * arm64
* windows
  * amd64
  * arm64

This will ensure that your provider supports the vast majority of potential users.

#### SHA256SUMS File
There must be an asset attached to the release with the following name:
```
terraform-provider-{{ provider name }}_{{ semver }}_SHA256SUMS
```

This file must be a plain text file with the following structure:

```
{{ compressed artifact sha256 checksum }}  {{ name of compressed binary artifact }}
```

*Note*: There must be two (2) space (dec. `32`) characters between the hash and the artifact name.

Example contents:
```
c4af2482975e2c5d253b0e78e515d063e59c93b3989780d8679b66828cb3e87a  terraform-provider-myprovider_0.1.0_darwin_amd64.zip
e6434014e9900ac13eb58b5d1ef5bc887bd66d2a272ae40984fe75869fa871f1  terraform-provider-myprovider_0.1.0_darwin_arm64.zip
ead99a92c724050c89a5ead1974e35ec31ffe90020e4fd47d95fbc8bed11ec6c  terraform-provider-myprovider_0.1.0_freebsd_amd64.zip
4fefa1161909c387475f1ce9476e04d7c0e6c25fdf84df2931de0633cd61aeaf  terraform-provider-myprovider_0.1.0_freebsd_arm64.zip
f95b79211d6753a350b0bcc527b8b5f33ac9ed74284d9359b107f8c2769710d2  terraform-provider-myprovider_0.1.0_linux_amd64.zip
22a801727b7e0b388099d8cf1c33537d6e0f7a85a6b1b233edaba41d48892455  terraform-provider-myprovider_0.1.0_linux_arm64.zip
4385047a17f29576213abaca2653bc32329bb72dbc3bddf937d2f9408043be89  terraform-provider-myprovider_0.1.0_windows_amd64.zip
1efe80f92155b70f124a0a55b238ded655ce464f06f69f18cc9230f72c5ac0e2  terraform-provider-myprovider_0.1.0_windows_arm64.zip
```

#### SHA256SUMS.sig File
There must be an asset attached to the release with the following name:
```
terraform-provider-{{ provider name }}_{{ semver }}_SHA256SUMS.sig
```

This file must be a gpg blob signature of the `SHA256SUMS` file using the GPG key you created earlier.

## 5. Action Configuration
This action is configured via environment variables.

### Environment Variables

| Name                      | Purpose                                                                                                           | Required | Default                      |
|---------------------------|-------------------------------------------------------------------------------------------------------------------|----------|------------------------------|
| `GITHUB_TOKEN`            | Github API token. This is created automatically when run and is accessible using `${{ secrets.GITHUB_TOKEN }}`    | yes      |                              |
| `GITHUB_REF_NAME`         | Automatically provided by [Github](https://docs.github.com/en/actions/learn-github-actions/environment-variables) | yes      |                              |
| `GITHUB_REPOSITORY`       | Automatically provided by [Github](https://docs.github.com/en/actions/learn-github-actions/environment-variables) | yes      |                              |
| `GITHUB_REPOSITORY_OWNER` | Automatically provided by [Github](https://docs.github.com/en/actions/learn-github-actions/environment-variables) | yes      |                              |
| `GITHUB_REQUEST_TTL`      | Maximum TTL for Github API requests                                                                               | no       | `"5s"`                       |
| `GITHUB_DOWNLOAD_TTL`     | Maximum TTL for Github release asset download requests                                                            | no       | `"5m"`                       |
| `TF_ADDRESS`              | Terraform cloud address                                                                                           | no       | `"https://app.terraform.io"` |
| `TF_TOKEN`                | Robot API token created earlier                                                                                   | yes      |                              |
| `TF_GPG_KEY_ID`           | Value from `key-id` field returned when registering your GPG key with Terraform Cloud                             | yes      |                              |
| `TF_REGISTRY_NAME`        | Name of registry to push provider to                                                                              | no       | `"private"`                  |
| `TF_ORGANIZATION_NAME`    | Name of your Terraform organization                                                                               | yes      |                              |
| `TF_NAMESPACE`            | Namespace for Provider.                                                                                           | yes      |                              |
| `TF_PROVIDER_NAME`        | Name of your provider.  Must match binary name prefix exactly.                                                    | yes      |                              |
| `TF_PROVIDER_PLATFORMS`   | Comma-separate list of versions supported by your provider.                                                       | no       | `"6.0"`                      |
| `TF_REQUEST_TTL`          | Maximum TTL for Terraform Cloud API requests                                                                      | no       | `"5s"`                       |
| `TF_UPLOAD_TTL`           | Maximum TTL for Terraform Cloud artifact uploads (including binaries)                                             | no       | `"5m"`                       |

### Example Config

```yaml
name: Release

# This causes the job only to run when a new tag with a prefix of `v` is pushed to github
on:
  push:
    tags:
      - 'v*'

jobs:
  create-release:
    runs-on: ubuntu-latest
    steps:
      # you may perform whatever release artifact creation steps you like here
      
      - uses: dcarbone/tfcloud-provider-push-action@v0.1.0 # version should be latest release
        if: ${{ success() }} # only run if previous steps succeeded
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }} # this is created for you by Github
          TF_TOKEN: ${{ secrets.TFCLOUD_API_KEY }} # this assumes you've created an Action secret with this name
          TF_GPG_KEY_ID: ${{ secrets.TFCLOUD_GPG_KEY_ID }} # this assumes you've created an Action secret with this name
          TF_REGISTRY_NAME: private
          TF_ORGANIZATION_NAME: myorg
          TF_NAMESPACE: myorg
          TF_PROVIDER_NAME: myprovider
```
