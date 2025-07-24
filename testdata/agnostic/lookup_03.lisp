(module m1)
;;
(defcolumns (X :i160))
(deflookup l1 (m2.Y) (X))

(module m2)
(defcolumns (Y :i256))
