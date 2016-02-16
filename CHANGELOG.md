# Changelog

## 0.2.1
* Feature - Update ECS-optimized AMIs to latest 2015.09.f

## 0.2.0
* Feature - Add support for new docker options in compose yaml file.
* Feature - Add new options to ecs-cli up (--image-id, --debug or --verbose).
* Feature - Add support for m4, d2, g2 instance types.
* Feature - Add new regions eu-central-1 and ap-southeast-1.
* Feature - Update ECS-optimized AMIs to latest 2015.09.e
  (with Amazon ECR support).
* Enhancement - Better error messaging for ecs-cli up and
  RegisterTaskDefinition API.
* Bug - Include region, account in key for local Task Definition cache.
* Bug - Change ordering of AWS Credential resolution for the ecs-cli.
* Bug - Minor bug fixes to CFN template (remove additional parameter from
  autoscaling creation, add internet gateway attachment dependency to public
  route)
