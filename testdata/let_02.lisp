(defpurefun ((vanishes! :@loob) x) x)

(defcolumns (A :@loob) B)
(defconstraint c1 ()
  (let ((C (* 1 B)))
    (if A
        (vanishes! 0)
        (vanishes! C))))
