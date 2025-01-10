;;error:6:9-12:malformed let assignment
(defpurefun ((vanishes! :@loob) x) x)
(defcolumns (A :@loob) B)

(defconstraint c1 ()
  (let ((C))
    (if A
        (vanishes! C))))
