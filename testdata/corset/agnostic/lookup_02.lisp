(module m1)
;;
(defcolumns (X_LO :i128) (X_HI :i128))
;;
(deflookup l1 (m2.Y) ((:: X_HI X_LO)))

(module m2)
(defcolumns (Y :i256))
