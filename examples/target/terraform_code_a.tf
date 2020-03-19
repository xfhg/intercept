module "ExampleBackEndApp8080" {

  source = "git::https://scm.yourcompany.xxx/modules/aws-alb-backend.git"

  subnet_ids          = ["${data.aws_subnet_ids.tier2_app_layer.ids}"]
  app_lb_listener_arn = "${module.example_alb_v2.aws_alb_listener_arn}"
  frontend_sg_id      = "${module.example_alb_v2.aws_alb_frontend_sg_id}"
  target_port         = "8080"
  target_protocol     = "HTTP"
  app_name            = "ExampleApp8080"
  vpc_id              = "${data.aws_vpc.vpc.id}"

  environment         = "${var.environment}"
  listener_path       = "/8080/*"
  desired_capacity    = 1
  max_size            = 1
  min_size            = 1
  ec2_keypair_name    = "${aws_key_pair.ec2_keypair.key_name}"
  instance_type       = "t3.medium"
  aws_apps_ami        = "${data.aws_ami.rhel7_ami.image_id}"
  user_data           = "${data.template_file.bootstrap_file_8080.rendered}"

  tag_map = "${
    merge(
      map("Name", "ExampleApp8080"),
      local.tag_map
    )
  }"
}