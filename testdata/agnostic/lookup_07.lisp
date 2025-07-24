(module m1)
;;
(defcolumns (P :i1) (X :i128))
(deflookup l1 (m2.Q m2.Y) (P X))

(module m2)
(defcolumns (Q :i256) (Y :i128))
