(defpurefun ((eq! :𝔽@loob) x y) (- x y))

(defcolumns (P1 :binary) (P2 :binary) (K1 :i16) (V1 :i16) (K2 :i16) (R :i16))
(defcomputed (V2) (map-if P2 K2 P1 K1 V1))
(defconstraint c1 () (eq! V2 R))
