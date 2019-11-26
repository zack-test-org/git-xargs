output "url" {
  value       = "http://${aws_route53_record.alias.name}"
  description = "The URL of the web service"
}
