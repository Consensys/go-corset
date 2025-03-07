;;error:6:15-18:malformed let assignment
(defpurefun ((vanishes! :𝔽@loob) x) x)
(defcolumns (A :i16@loob) B)

(defconstraint c1 ()
  (let ((C B) (D))
    (if A
        (vanishes! C))))
