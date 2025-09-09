(module m1)
;;
(defcolumns (X :i256))
(deflookup (l1 :unchecked) (m2.Y) (X))

(module m2)
(defcolumns (Y :i128))
