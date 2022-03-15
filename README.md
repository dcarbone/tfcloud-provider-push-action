# tfcloud-provider-push-action
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
makes sense for your team.

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
