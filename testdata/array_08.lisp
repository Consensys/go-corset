(defpurefun ((vanishes! :𝔽@loob :force) e0) e0)

(module m1)

(defcolumns (SEL :binary))

(defperspective test
  ;; Selector
  SEL
  ;; Columns
  ((BYTE :byte :array [2])))

(defun (hi)
  [test/BYTE 1])

(defconstraint check (:perspective test)
  (vanishes! (hi)))
