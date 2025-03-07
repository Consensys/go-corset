(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns STAMP X)
;; STAMP == 0 || X == 1 || X == 2
(defconstraint c1 (:guard STAMP)
  (vanishes! (* (- X 1) (- X 2))))
