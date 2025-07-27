BEGIN {
  if "1" == "1" {
    print "hello"
  }

  if "1" != "1" {
    print "hello"
  } else {
    print "else"
  }

  if "1" != "1" {
    print "hello"
  } else if "1" != 2 {
    print "else if"
  } else {
    print "else"
  }
}
