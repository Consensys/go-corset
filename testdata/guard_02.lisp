(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns ST (A :i16@loob) B)
;; STAMP == 0 || X == 1 || X == 2
(defconstraint c1 (:guard ST)
  (if A
      (vanishes! B)))
