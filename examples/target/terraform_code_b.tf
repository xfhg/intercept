module "my_stateless_web_app" {
  source = "https://artifactory.yourcompany.xxx/modules/elb-module.zip"

  region_name                 = "${var.region}"
  environment                 = "${var.environment}"
  user_data                   = "${data.template_file.bootstrap_file.rendered}"
  instance_type               = "t3.medium"
  min_size                    = "2"
  max_size                    = "2"
  desired_capacity            = "2"

  aws_apps_ami                = "${data.aws_ami.rhel7_ami.id}"

  sg_app_port                 = "${var.sg_app_port}"
  default_lb_protocol         = "${var.default_lb_protocol}"
  default_elb_app_port        = "${var.default_elb_app_port}"
  default_healthcheck_target  = "${var.default_healthcheck_target}"

  elb_listener                = [
    {
      instance_port       = "${var.instance_port}"
      instance_protocol   = "${var.instance_protocol}"
      lb_port             = "${var.lb_port}"
      lb_protocol         = "${var.lb_protocol}"
      ssl_certificate_id  = "${var.ssl_certificate}"
    }
  ]

  elb_health_check            = [
    {
      target              = "${var.target}"
      interval            = "${var.interval}"
      healthy_threshold   = "${var.healthy_threshold}"
      unhealthy_threshold = "${var.unhealthy_threshold}"
      timeout             = "${var.timeout}"
    }
  ]

  app_name                = "elb-testing"
  ec2_keypair_name        = "${var.ec2_keypair_name}"

  vpc_zone_identifier     = "${data.aws_subnet_ids.tier2_app_layer.ids}"
  vpc_id                  = "${data.aws_vpc.vpc.id}"
  subnet_ids              = "${data.aws_subnet_ids.tier2_app_layer.ids}"


  ###  Parameters block below are optional - remove if unused  ###

  root_volume_size         = 20
  enable_launch_template   = "true"
  enable_stickiness        = "false"
  cookie_expiration_period = 600
  enable_nightly_shutdown  = "false"
  business_hours_start     = "0 22 * * *"
  business_hours_end       = "0 14 * * *"

  

}