(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (A :i16@loob) B)
(defconstraint c1 ()
  (let ((C (* 1 B)))
    (if A
        (vanishes! 0)
        (vanishes! C))))
