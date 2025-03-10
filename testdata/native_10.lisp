(defpurefun ((eq! :𝔽@loob) x y) (- x y))

(defcolumns (P :binary) (Q :binary) (X :i8) (Y :i16))
(defcomputed (Z) (fwd-fill-within P Q X))
(defconstraint c1 () (eq! Y Z))
