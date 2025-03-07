(defpurefun ((eq! :𝔽@loob) x y) (- x y))

(defcolumns (P :binary) (Q :binary) (X :i16) Y)
(defcomputed (Z) (bwd-fill-within P Q X))
(defconstraint c1 () (eq! Y Z))
