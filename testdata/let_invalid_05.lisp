;;error:6:16-17:already defined
(defpurefun ((vanishes! :@loob) x) x)
(defcolumns (A :@loob) B)

(defconstraint c1 ()
  (let ((C B) (C B))
    (if A
        (vanishes! C))))
