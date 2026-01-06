(module sq)
(defcolumns (X :i16) (Y :i16))
;; Y = (X * X) % 65536
(deflookup l1 (mul.arg1 mul.arg2 mul.res) (X X Y))
