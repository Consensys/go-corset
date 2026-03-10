(module m1)
(defcolumns (X :i32) (Y :i32))
(defclookup (l1 :unchecked) 1 (cast.x cast.y) 1 (X Y))
