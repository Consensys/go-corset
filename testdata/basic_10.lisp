(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16) (Y :i32))
;; Y == X*X
(defconstraint c1 () (vanishes! (- Y (^ X 2))))
