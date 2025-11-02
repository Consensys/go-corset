(defcolumns (X :i256) (Y :i256) (Z :i256))
;; X == Y + Z
(defconstraint c1 () (== (+ X Y) Z))
