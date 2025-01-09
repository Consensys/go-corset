;;error:6:15-18:malformed let assignment
(defpurefun ((vanishes! :@loob) x) x)
(defcolumns (A :@loob) B)

(defconstraint c1 ()
  (let ((C B) (D))
    (if A
        (vanishes! C))))
