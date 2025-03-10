(defpurefun ((eq! :𝔽@loob) x y) (- x y))

(defcolumns (P :binary) (K1 :i16) (V1 :i16) (K2 :i16) (R :i16))
(defcomputed (V2) (map-if P K2 P K1 V1))
(defconstraint c1 () (eq! V2 R))
