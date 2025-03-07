(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns X Y)
;; Y == X*X
(defconstraint c1 () (vanishes! (- Y (^ X 2))))
