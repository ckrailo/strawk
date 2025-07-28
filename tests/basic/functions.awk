
function foo(a,b) {
  print "foo"
  return "a"
}

BEGIN {
  x = foo(1,2)
  print x
}
