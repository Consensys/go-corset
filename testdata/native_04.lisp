(defpurefun ((eq! :𝔽@loob) x y) (- x y))

(defcolumns P X Y)
(defcomputed (Z) (fwd-unchanged-within P X))
(defconstraint c1 () (eq! Y Z))
