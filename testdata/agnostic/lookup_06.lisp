(module m1)
;;
(defcolumns (P :i1) (X :i128))
(defclookup l1 (m2.Q m2.Y) P (1 X))

(module m2)
(defcolumns (Q :i256) (Y :i128))
