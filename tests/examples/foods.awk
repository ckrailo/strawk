# If a regex is matched, the full value is put in $0 and each capture group in $1, $2...
/Name: (.*)\nFavorite Food: (.*)\n?/ {
   if ($2 in foods) {
     foods[$2]++
   } else {
     foods[$2] = 1
   }
}

END {
   count, mostpopular = 0, ""
   for (food in foods) {
     if foods[food] > count {
       count = foods[food]
       mostpopular = food
     }
   }
   print "The most popular food is:", mostpopular
}
