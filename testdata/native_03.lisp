(defpurefun ((eq! :𝔽@loob) x y) (- x y))

(defcolumns (P :i2) (X :i16) (Y :i16))
(defcomputed (Z) (bwd-changes-within P X))
(defconstraint c1 () (eq! Y Z))
