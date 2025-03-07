(defpurefun ((vanishes! :𝔽@loob) x) x)

(defconst TWO 2)
(defcolumns (X :i16) (Y :i16))
;; Y == X*X
(defconstraint c1 () (vanishes! (- Y (^ X TWO))))
