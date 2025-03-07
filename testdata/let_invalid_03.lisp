;;error:6:9-12:malformed let assignment
(defpurefun ((vanishes! :𝔽@loob) x) x)
(defcolumns (A :i16@loob) B)

(defconstraint c1 ()
  (let ((C))
    (if A
        (vanishes! C))))
