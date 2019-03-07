resource "google_storage_bucket" "image-store" {
  name     = "deploy-dev-test-buckets"
}

resource "google_storage_bucket_object" "picture1" {
  name   = "cat2"
  source = "./images/cat1.jpeg"
  bucket = "${google_storage_bucket.image-store.name}"
}

/*
resource "google_storage_bucket_object" "picture2" {
  name   = "cat2"
  source = "/images/cat2.jpeg"
  bucket = "${google_storage_bucket.image-store.name}"
}*/
/*data "google_storage_bucket_object" "picture" {
  name   = "${google_storage_bucket_object.picture1.name}"
  bucket = "${google_storage_bucket.image-store.name}"
}*/