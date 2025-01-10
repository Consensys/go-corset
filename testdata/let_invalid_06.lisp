;;error:6:12-13:unknown symbol
(defpurefun ((vanishes! :@loob) x) x)
(defcolumns (A :@loob) B)

(defconstraint c1 ()
  (let ((B C))
    (if A
        (vanishes! B))))
